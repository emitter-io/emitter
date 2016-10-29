#region Copyright (c) 2009-2016 Misakai Ltd.
/*************************************************************************
* This program is free software: you can redistribute it and/or modify
* it under the terms of the GNU Affero General Public License as
* published by the Free Software Foundation, either version 3 of the
* License, or(at your option) any later version.
*
* This program is distributed in the hope that it will be useful,
* but WITHOUT ANY WARRANTY; without even the implied warranty of
*  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.See the
* GNU Affero General Public License for more details.
*
* You should have received a copy of the GNU Affero General Public License
* along with this program.If not, see<http://www.gnu.org/licenses/>.
*************************************************************************/
#endregion Copyright (c) 2009-2016 Misakai Ltd.

using System;
using System.Security.Cryptography.X509Certificates;

namespace Emitter.Providers
{
    /// <summary>
    /// Represents a certificate provider which loads the certificate from the Vault storage.
    /// </summary>
    public sealed class VaultCertificateProvider : CertificateProvider
    {
        private object Lock = new object();
        private bool X509Resolved = false;
        private X509Certificate2 X509 = null;

        /// <summary>
        /// Gets or sets the certificate to use for TLS encryption/decryption.
        /// </summary>
        public override X509Certificate2 Certificate
        {
            get
            {
                lock (this.Lock)
                {
                    if (!this.X509Resolved && this.X509 == null)
                        this.X509 = this.AutoResolveCertificate();
                    return this.X509;
                }
            }
            set
            {
                lock (this.Lock)
                {
                    this.X509 = value;
                }
            }
        }

        #region Private Members

        /// <summary>
        /// Attempts to auto-resolve and load the certificate from the data folder.
        /// </summary>
        private X509Certificate2 AutoResolveCertificate()
        {
            this.X509Resolved = true;
            try
            {
                // Get the certificate from the broker
                var secret = Service.Providers
                    .Resolve<SecurityProvider>()
                    .GetSecret("emitter/cert/broker");

                // Load the certificate from raw data
                var x509 = new X509Certificate2(Convert.FromBase64String(secret), this.Password);

                // Construct a message
                var message = String.Format("X509: {0} ({1}, exp. {2}) {3}",
                    x509.FriendlyName,
                    x509.SignatureAlgorithm.FriendlyName,
                    x509.NotAfter.ToString("dd/MM/yyyy"),
                    x509.HasPrivateKey ? "with private key" : String.Empty
                    );

                Service.Logger.Log(LogLevel.Info, message);
                return x509;
            }
            catch (Exception ex)
            {
                // Failed
                Service.Logger.Log(ex);
                return null;
            }
        }

        #endregion Private Members
    }
}